package main

func contains(accounts map[string]bool, account string) bool {
	_, ok := accounts[account]
	return ok
}
