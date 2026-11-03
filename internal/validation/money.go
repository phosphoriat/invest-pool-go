package validation

const MaxMoney int64 = 9_000_000_000_000_000

// Money Check that an amount is positive and within the supported minor-unit range
func Money(amount int64) bool {
	return amount > 0 && amount <= MaxMoney
}

// OperationKey Check that an idempotency key contains between 1 and 128 bytes
func OperationKey(key string) bool {
	return len(key) >= 1 && len(key) <= 128
}
