package main

// Run some action, with some particular action tag on top of outputs
// printed by that command. Recover panics, log them, and exit. The
// given callback function is provided a tagged logger to use for its
// output.
func RunRecoverWithTag(action string, output Output, cb func(Output) error) {
	output.Debug("starting %s\n", action)
	tagged := TaggedOutput{output, action}
	defer func() {
		if r := recover(); r != nil {
			tagged.Error("panic: %s\n", r)
			output.Exit(2)
		}
	}()
	err := cb(tagged)
	if err != nil {
		tagged.Error("%s\n", err)
		output.Exit(1)
	}
}
