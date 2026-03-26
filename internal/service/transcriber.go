package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	openai "github.com/sashabaranov/go-openai"

	"github.com/hnam/notafly/internal/config"
	"github.com/hnam/notafly/internal/model"
	"go.uber.org/zap"
)

// System prompts — ported verbatim from Python speech_to_text.py lines 59-123.
const (
	summaryPrompt = "You are a highly skilled AI trained in language comprehension and summarization. I would like you to read the following text and summarize it into a concise abstract paragraph. Aim to retain the most important points, providing a coherent and readable summary that could help a person understand the main points of the discussion without needing to read the entire text. Please avoid unnecessary details or tangential points."

	keyPointsPrompt = "You are a proficient AI with a specialty in distilling information into key points. Based on the following text, identify and list the main points that were discussed or brought up. These should be the most important ideas, findings, or topics that are crucial to the essence of the discussion. Your goal is to provide a list that someone could read to quickly understand what was talked about."

	actionItemsPrompt = "You are an AI expert in analyzing conversations and extracting action items. Please review the text and identify any tasks, assignments, or actions that were agreed upon or mentioned as needing to be done. These could be tasks assigned to specific individuals, or general actions that the group has decided to take. Please list these action items clearly and concisely."

	sentimentPrompt = "As an AI with expertise in language and emotion analysis, your task is to analyze the sentiment of the following text. Please consider the overall tone of the discussion, the emotion conveyed by the language used, and the context in which words and phrases are used. Indicate whether the sentiment is generally positive, negative, or neutral, and provide brief explanations for your analysis where possible."
)

type TranscriberService struct {
	client *openai.Client
	config *config.Config
	logger *zap.Logger
}

func NewTranscriberService(cfg *config.Config, logger *zap.Logger) *TranscriberService {
	client := openai.NewClient(cfg.OpenAIAPIKey)
	return &TranscriberService{
		client: client,
		config: cfg,
		logger: logger,
	}
}

// Transcribe sends an audio file to Whisper API and returns the transcribed text.
func (s *TranscriberService) Transcribe(ctx context.Context, audioPath string) (string, error) {
	s.logger.Info("transcribing audio", zap.String("path", audioPath))

	resp, err := s.client.CreateTranslation(ctx, openai.AudioRequest{
		Model:    s.config.WhisperModel,
		FilePath: audioPath,
	})
	if err != nil {
		return "", fmt.Errorf("whisper transcription failed: %w", err)
	}

	s.logger.Info("transcription complete", zap.Int("length", len(resp.Text)))
	return resp.Text, nil
}

// AnalyzeMeeting runs 4 GPT analyses CONCURRENTLY — 4x faster than Python's sequential approach.
func (s *TranscriberService) AnalyzeMeeting(ctx context.Context, transcription string) (*model.MeetingMinutes, error) {
	s.logger.Info("analyzing meeting transcript")

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		minutes model.MeetingMinutes
		errs    []error
	)

	analyses := []struct {
		name   string
		prompt string
		target *string
	}{
		{"summary", summaryPrompt, &minutes.AbstractSummary},
		{"key_points", keyPointsPrompt, &minutes.KeyPoints},
		{"action_items", actionItemsPrompt, &minutes.ActionItems},
		{"sentiment", sentimentPrompt, &minutes.Sentiment},
	}

	for _, a := range analyses {
		wg.Add(1)
		go func(name, prompt string, target *string) {
			defer wg.Done()

			result, err := s.chatCompletion(ctx, prompt, transcription)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", name, err))
				return
			}
			*target = result
			s.logger.Debug("analysis done", zap.String("type", name))
		}(a.name, a.prompt, a.target)
	}

	wg.Wait()

	if len(errs) > 0 {
		return nil, fmt.Errorf("analysis errors: %v", errs)
	}

	s.logger.Info("all analyses complete")
	return &minutes, nil
}

// TranscribeAndAnalyze is the full pipeline: compress → transcribe → analyze → save JSON.
func (s *TranscriberService) TranscribeAndAnalyze(ctx context.Context, audioPath string, recorder *RecorderService) (*model.TranscriptionResult, error) {
	// Compress if needed
	processedPath, err := recorder.CompressIfNeeded(audioPath)
	if err != nil {
		return nil, fmt.Errorf("compress audio: %w", err)
	}

	// Transcribe
	text, err := s.Transcribe(ctx, processedPath)
	if err != nil {
		return nil, err
	}

	// Analyze
	minutes, err := s.AnalyzeMeeting(ctx, text)
	if err != nil {
		return nil, err
	}

	result := &model.TranscriptionResult{
		Text:    text,
		Minutes: *minutes,
	}

	// Save to JSON
	if err := s.SaveToJSON(result); err != nil {
		s.logger.Warn("failed to save JSON", zap.Error(err))
		// Non-fatal — still return the result
	}

	return result, nil
}

// SaveToJSON writes the transcription result to a JSON file in the temp directory.
func (s *TranscriberService) SaveToJSON(result *model.TranscriptionResult) error {
	dir := os.TempDir()
	filename := fmt.Sprintf("meeting_data_%s.json", time.Now().Format("20060102150405"))
	path := filepath.Join(dir, filename)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create json file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	s.logger.Info("saved meeting data", zap.String("path", path))
	return nil
}

// chatCompletion sends a single chat completion request to GPT.
func (s *TranscriberService) chatCompletion(ctx context.Context, systemPrompt, userContent string) (string, error) {
	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       s.config.GPTModel,
		Temperature: 0,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userContent},
		},
	})
	if err != nil {
		return "", fmt.Errorf("chat completion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return resp.Choices[0].Message.Content, nil
}

// Prompts returns the system prompts for external testing/inspection.
func Prompts() map[string]string {
	return map[string]string{
		"summary":      summaryPrompt,
		"key_points":   keyPointsPrompt,
		"action_items": actionItemsPrompt,
		"sentiment":    sentimentPrompt,
	}
}
