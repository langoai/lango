package knowledge

import "errors"

var (
	ErrKnowledgeNotFound = errors.New("knowledge not found")
	ErrLearningNotFound  = errors.New("learning not found")
)
