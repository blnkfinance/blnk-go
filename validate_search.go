package blnkgo

import "fmt"

// ValidateStartReindexRequest performs client-side checks before POST /search/reindex.
func ValidateStartReindexRequest(body StartReindexRequest) error {
	if body.BatchSize != nil && *body.BatchSize < 1 {
		return fmt.Errorf("batch_size must be a positive integer if provided")
	}
	return nil
}
