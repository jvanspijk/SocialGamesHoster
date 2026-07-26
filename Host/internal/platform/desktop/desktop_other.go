//go:build !windows

package desktop

func Supported() bool {
	return false
}

func AcquireSingleInstance(string, bool) (bool, func(), error) {
	return true, func() {}, nil
}

func OpenURL(string) error {
	return nil
}

func CopyText(string) error {
	return nil
}

func UpdateFirewallPort(int) error {
	return nil
}

func ShowError(string, string) {}

func Run(Actions) error {
	return nil
}
