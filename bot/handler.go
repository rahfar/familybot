package bot

func ping(message string) string {
	if (message == "ping") || (message == "пинг") {
		return "pong"
	}
	return ""
}
