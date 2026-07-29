package repo

import (
	"fmt"
	"matching_service/ent"
	cerr "matching_service/internal/common/errors"
)

func wrapError(err error) error {
	if err == nil {
		return nil
	}

	switch err.(type) {
	case *ent.NotFoundError:
		return fmt.Errorf("%w: %v", cerr.ErrRecordNotFound, err)
	case *ent.ConstraintError:
		return fmt.Errorf("%w: %v", cerr.ErrInlavidInput)
	case *ent.ValidationError:
		return fmt.Errorf("%w: %v", cerr.ErrValidationFailed, err)
	default:
		return fmt.Errorf("%w: db execution failed - %v", cerr.ErrInternalServer, err)
	}
}
