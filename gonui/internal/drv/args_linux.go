package drv

// Args — аргументы командной строки из /proc/self/cmdline (байты как есть): у
// библиотеки нет пакета os (ADR-7).
func Args() []string {
	b, ok := ReadFile("/proc/self/cmdline")
	if !ok {
		return nil
	}
	var args []string
	for len(b) > 0 {
		i := 0
		for i < len(b) && b[i] != 0 {
			i++
		}
		args = append(args, string(b[:i]))
		b = b[min(i+1, len(b)):]
	}
	return args
}
