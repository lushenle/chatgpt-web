package api

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	gogpt "github.com/sashabaranov/go-openai"
)

func (server *Server) completion(ctx *gin.Context) {
	var request gogpt.ChatCompletionRequest
	err := ctx.BindJSON(&request)
	if err != nil {
		server.responseJson(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	log.Println(request)
	if request.Messages == nil {
		server.responseJson(ctx, http.StatusBadRequest, "request messages required", nil)
		return
	}

	gptConfig := gogpt.DefaultConfig(server.config.ChatGPT.ChatGPTAPIKey)

	if server.config.ChatGPT.Proxy != "" {
		// creates http transport object, sets proxy
		proxyUrl, err := url.Parse(server.config.ChatGPT.Proxy)
		if err != nil {
			log.Fatalf("parse proxy err: %v", err)
		}
		transport := &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
		gptConfig.HTTPClient = &http.Client{Transport: transport}
	}

	client := gogpt.NewClientWithConfig(gptConfig)

	if strings.EqualFold(request.Messages[0].Role, "system") {
		newMessage := append([]gogpt.ChatCompletionMessage{{Role: "system",
			Content: "You're an AI assistant, and I need you to simulate a programmer to answer my questions."}}, request.Messages...)
		request.Messages = newMessage
	}
	log.Println(request.Messages)

	if strings.EqualFold(server.config.ChatGPT.Model, gogpt.GPT3Dot5Turbo) ||
		strings.EqualFold(server.config.ChatGPT.Model, gogpt.GPT3Dot5Turbo0301) {
		request.Model = server.config.ChatGPT.Model
		resp, err := client.CreateChatCompletion(ctx, request)
		if err != nil {
			server.responseJson(ctx, http.StatusInternalServerError, err.Error(), nil)
			return
		}
		server.responseJson(ctx, http.StatusOK, "", gin.H{
			"reply":    resp.Choices[0].Message.Content,
			"messages": append(request.Messages, resp.Choices[0].Message),
		})
	} else {
		prompt := ""
		for _, item := range request.Messages {
			prompt += item.Content + "/n"
		}
		prompt = strings.Trim(prompt, "/n")
		log.Printf("request prompt: %v\n", prompt)

		req := gogpt.CompletionRequest{
			Model:            server.config.ChatGPT.Model,
			MaxTokens:        server.config.ChatGPT.MaxTokens,
			TopP:             1,
			FrequencyPenalty: 0.9,
			PresencePenalty:  0.9,
			Prompt:           prompt,
		}
		resp, err := client.CreateCompletion(ctx, req)
		if err != nil {
			server.responseJson(ctx, http.StatusInternalServerError, err.Error(), nil)
			return
		}

		server.responseJson(ctx, http.StatusOK, "", gin.H{
			"reply": resp.Choices[0].Text,
			"messages": append(request.Messages, gogpt.ChatCompletionMessage{
				Role:    "assistant",
				Content: resp.Choices[0].Text,
			}),
		})
	}
}
