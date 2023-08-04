package openweathermap

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/mfmayer/gosk"
	"github.com/mfmayer/gosk/pkg/gpt"
	"github.com/mfmayer/gosk/pkg/llm"
	"github.com/tidwall/gjson"
)

//go:embed assets
var fsAssets embed.FS

func init() {
	godotenv.Load()
}

// GetOpenAIKey tries to retrieve OpenAI key from "OPENAI_API_KEY" environment variable or .env file in current workgin directory
func getAPIKey() (string, error) {
	key := os.Getenv("OPENWEATHERMAP_API_KEY")
	if key == "" {
		return "", errors.New("OPENWEATHERMAP_API_KEY not set")
	}
	return key, nil
}

func getGenerators(generatorConfigs map[string]llm.GeneratorConfig) (generators llm.GeneratorMap, err error) {
	generators = llm.GeneratorMap{}
	for k, v := range generatorConfigs {
		switch v.Model() {
		case "gpt-3.5-turbo":
			generator, err := gpt.NewGenerator(gpt.WithConfig(v))
			if err != nil {
				return nil, err
			}
			generators[k] = generator
			continue
		}
		return nil, llm.ErrUnknownGeneratorModel
	}
	return
}

// RESTGetRequest führt eine HTTP-GET-Anfrage an eine bestimmte URL durch und gibt das Ergebnis als gjson.Result zurück.
func RESTGetRequest(url string) (gjson.Result, error) {
	resp, err := http.Get(url)
	if err != nil {
		return gjson.Result{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return gjson.Result{}, err
	}

	result := gjson.ParseBytes(body)

	return result, nil
}

func New() (skill *gosk.Skill, err error) {
	subFS, err := fs.Sub(fsAssets, "assets")
	if err != nil {
		return
	}
	skill, err = gosk.ParseSemanticSkillFromFS(subFS, getGenerators)
	if err != nil {
		return
	}

	getWeatherData := &gosk.Function{
		Name:        "getWeatherData",
		Description: "Get weather data for a given location.",
		Parameters: map[string]*gosk.Parameter{
			"location.latitude": {
				Description: "The location's geocoordinates latitude.",
				Type:        gosk.TypeNumber,
				Required:    true,
			},
			"location.longitude": {
				Description: "The location's geocoordinates longitude.",
				Type:        gosk.TypeNumber,
				Required:    true,
			},
		},
		Plannable: true,
		Call: func(input llm.Content) (output llm.Content, err error) {
			apiKey, err := getAPIKey()
			if err != nil {
				return nil, err
			}
			lat, ok := input.Property("location.latitude").Value().(float64)
			if !ok {
				return nil, errors.New("location.latitude must be a number")
			}
			lon, ok := input.Property("location.longitude").Value().(float64)
			if !ok {
				return nil, errors.New("location.longitude must be a number")
			}
			url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?lat=%f&lon=%f&appid=%s", lat, lon, apiKey)
			result, err := RESTGetRequest(url)
			if err != nil {
				return nil, err
			}
			output = llm.NewContent()
			output.With("weather.description", result.Get("list.0.weather.0.description").String())
			output.With("weather.time", time.Unix(result.Get("list.0.dt").Int(), 0).String())
			return
		},
	}
	skill.Functions["getWeatherData"] = getWeatherData
	return
}
