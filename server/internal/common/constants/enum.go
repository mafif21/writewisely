package constants

// TaskType is the IELTS writing task (data-design §3, writing_submission.task_type).
type TaskType string

// TaskType values.
const (
	TaskTypeTask1Academic TaskType = "task1_academic"
	TaskTypeTask2         TaskType = "task2"
)

// PromptSource says whether the user supplied the prompt or AI generated it (writing_submission.prompt_source).
type PromptSource string

// PromptSource values.
const (
	PromptSourceUserProvided PromptSource = "user_provided"
	PromptSourceAIGenerated  PromptSource = "ai_generated"
)

// ChartMode is how a Task 1 chart is supplied (writing_submission.chart_mode).
type ChartMode string

// ChartMode values.
const (
	ChartModeNone          ChartMode = "none"
	ChartModeUploadedImage ChartMode = "uploaded_image"
	ChartModeGeneratedData ChartMode = "generated_data"
)
