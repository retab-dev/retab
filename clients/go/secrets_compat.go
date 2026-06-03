// @oagen-ignore-file
package retab

import "context"

// GetValue retrieves a decrypted secret value.
func (s *SecretService) GetValue(ctx context.Context, name string, opts ...RequestOption) (*SecretValueResponse, error) {
	return s.ListValue(ctx, name, opts...)
}
