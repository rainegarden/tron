package terminal

type OutputMsg struct {
	Line string
}

type CommandFinishedMsg struct {
	Command  string
	ExitCode int
	Err      error
}

type CommandStartedMsg struct {
	Command string
	Cwd     string
}
