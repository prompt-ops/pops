package shell

type commandMsg struct {
	command string
}

type outputMsg struct {
	output string
}

type answerMsg struct {
	answer string
}

type checkPassedMsg struct {
}

type errMsg struct {
	err error
}
