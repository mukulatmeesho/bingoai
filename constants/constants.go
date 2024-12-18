package constants

const (
	DefaultOllamaURLGen  = "http://localhost:11434/api/generate"
	DefaultOllamaURLChat = "http://localhost:11434/api/chat"
	//DefaultModel         = "wizardlm2"
	DefaultModel = "llama3.2"
	Prompt       = "Please provide brief and concise responses. Please don't do extra other formating "

	HistoryDir = "chatbot/history/files"
)

//
//viper.SetDefault("model", "llama3.2")
//viper.SetDefault("temperature", 0.5)
