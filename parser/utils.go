package parser

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"librus/helper"
	"log"
)

func logAction(name string) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		if helper.IsDebug() {
			log.Printf("Операция: %s", name)
		}
		return nil
	})
}

func printLog(text string) {
	if helper.IsDebug() {
		fmt.Println(text)
	}
}

func removeDuplicates(strings []string) []string {
	encountered := map[string]struct{}{}
	var result []string

	for _, str := range strings {
		if _, ok := encountered[str]; !ok {
			encountered[str], result = struct{}{}, append(result, str)
		}
	}

	return result
}
