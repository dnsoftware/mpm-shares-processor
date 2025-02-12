package share

import (
	"strings"

	"github.com/dnsoftware/mpm-shares-processor/constants"
)

// WalletFromWorkerfull получить кошелек из полного имени воркера
func WalletFromWorkerfull(workerfull string) string {
	parts := strings.Split(workerfull, constants.WorkerSeparator)

	return parts[0]
}

// WorkerFromWorkerfull получить короткое имя воркера из полного имени воркера
func WorkerFromWorkerfull(workerfull string) string {
	parts := strings.Split(workerfull, constants.WorkerSeparator)

	if len(parts) > 1 {
		return parts[1]
	} else {
		return ""
	}

}
