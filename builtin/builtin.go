// Package builtin — встроенные данные по умолчанию: план 24 недель и банк вопросов.
// plan.json лежит рядом с кодом и вшивается в бинарник через go:embed;
// пользователь может заменить его своим файлом plan.json в каталоге данных
// или править прямо в приложении (Настройки → План).
package builtin

import (
	_ "embed"
	"encoding/json"

	"studytracker/model"
)

//go:embed plan.json
var PlanJSON []byte

//go:embed questions.json
var QuestionsJSON []byte

// DefaultPlan — разбор встроенного плана. Паникует только при встроенном
// битом plan.json — этого не случится: файл проверяется тестом TestEmbeddedPlan.
func DefaultPlan() *model.Plan {
	var p model.Plan
	if err := json.Unmarshal(PlanJSON, &p); err != nil {
		panic("builtin: битый plan.json: " + err.Error())
	}
	return &p
}

// DefaultQuestions — стартовый банк вопросов.
func DefaultQuestions() []*model.QuizQuestion {
	var qs []*model.QuizQuestion
	if err := json.Unmarshal(QuestionsJSON, &qs); err != nil {
		panic("builtin: битый questions.json: " + err.Error())
	}
	return qs
}
