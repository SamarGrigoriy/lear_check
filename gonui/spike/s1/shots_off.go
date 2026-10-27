//go:build !shots

package main

func shot(string) {
	if shotDir != "" {
		println("s1: снимки не делаются — соберите с -tags shots")
		shotDir = ""
	}
}
