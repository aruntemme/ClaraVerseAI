import { useState, useEffect, useCallback, useImperativeHandle, forwardRef, useRef } from 'react';
import { motion } from 'framer-motion';
import { AlertCircle, HelpCircle, CheckCircle } from 'lucide-react';
import type { ActivePrompt, PromptAnswer, ValidationResult } from '@/types/interactivePrompt';
import type { InteractiveQuestion } from '@/types/websocket';
import './InteractivePromptMessage.css';

// Handle interface for parent component (CommandCenter) to trigger submit
export interface PromptMessageHandle {
  handleSubmit: () => void;
  isValid: () => boolean;
}

interface InteractivePromptMessageProps {
  prompt: ActivePrompt;
  onSubmit: (answers: Record<string, PromptAnswer>) => void;
  onValidationChange?: (isValid: boolean) => void;
  isSubmitting?: boolean;
}

export const InteractivePromptMessage = forwardRef<
  PromptMessageHandle,
  InteractivePromptMessageProps
>(({ prompt, onSubmit, onValidationChange, isSubmitting = false }, ref) => {
  const [answers, setAnswers] = useState<Record<string, PromptAnswer>>({});
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [touchedFields, setTouchedFields] = useState<Set<string>>(new Set());
  const cardRef = useRef<HTMLDivElement>(null);

  // Initialize answers with default values
  useEffect(() => {
    const initialAnswers: Record<string, PromptAnswer> = {};
    prompt.questions.forEach(question => {
      if (question.default_value !== undefined) {
        initialAnswers[question.id] = {
          questionId: question.id,
          value: question.default_value,
          isOther: false,
        };
      }
    });
    setAnswers(initialAnswers);
    setErrors({});
    setTouchedFields(new Set());

    // Auto-scroll to the prompt when it appears
    setTimeout(() => {
      cardRef.current?.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    }, 100);
  }, [prompt]);

  // Validate a single answer
  const validateAnswer = useCallback(
    (question: InteractiveQuestion, answer: PromptAnswer | undefined): string | null => {
      // Required validation
      if (question.required) {
        if (!answer || answer.value === '' || answer.value === null || answer.value === undefined) {
          return 'This field is required';
        }

        // Check for empty arrays in multi-select
        if (Array.isArray(answer.value) && answer.value.length === 0) {
          return 'Please select at least one option';
        }
      }

      if (!answer || !answer.value) {
        return null; // Optional field, no validation needed
      }

      const { validation } = question;
      if (!validation) return null;

      // Type-specific validation
      if (question.type === 'text') {
        const textValue = String(answer.value);

        if (validation.min_length && textValue.length < validation.min_length) {
          return `Minimum length is ${validation.min_length} characters`;
        }

        if (validation.max_length && textValue.length > validation.max_length) {
          return `Maximum length is ${validation.max_length} characters`;
        }

        if (validation.pattern) {
          const regex = new RegExp(validation.pattern);
          if (!regex.test(textValue)) {
            return 'Invalid format';
          }
        }
      }

      if (question.type === 'number') {
        const numValue = Number(answer.value);

        if (validation.min !== undefined && numValue < validation.min) {
          return `Minimum value is ${validation.min}`;
        }

        if (validation.max !== undefined && numValue > validation.max) {
          return `Maximum value is ${validation.max}`;
        }
      }

      return null;
    },
    []
  );

  // Validate all answers
  const validateAll = useCallback((): ValidationResult => {
    const newErrors: Record<string, string> = {};

    prompt.questions.forEach(question => {
      const answer = answers[question.id];
      const error = validateAnswer(question, answer);
      if (error) {
        newErrors[question.id] = error;
      }
    });

    return {
      isValid: Object.keys(newErrors).length === 0,
      errors: newErrors,
    };
  }, [prompt.questions, answers, validateAnswer]);

  // Track validation changes whenever answers change
  useEffect(() => {
    const validation = validateAll();
    console.log('🔍 Validation Update:', {
      answers,
      validation,
      errors: validation.errors,
      questionCount: prompt.questions.length,
      answeredCount: Object.keys(answers).length,
    });
    console.log(
      '📞 Calling onValidationChange with:',
      validation.isValid,
      'callback exists:',
      !!onValidationChange
    );
    onValidationChange?.(validation.isValid);
  }, [answers, validateAll, onValidationChange, prompt.questions.length]);

  // Handle answer change
  const handleAnswerChange = (
    questionId: string,
    value: PromptAnswer['value'],
    isOther: boolean = false
  ) => {
    setAnswers(prev => ({
      ...prev,
      [questionId]: {
        questionId,
        value,
        isOther,
      },
    }));

    // Clear error when user starts typing
    if (errors[questionId]) {
      setErrors(prev => {
        const newErrors = { ...prev };
        delete newErrors[questionId];
        return newErrors;
      });
    }
  };

  // Handle "Other" text input change
  const handleOtherTextChange = (questionId: string, otherText: string) => {
    setAnswers(prev => ({
      ...prev,
      [questionId]: {
        ...prev[questionId],
        questionId,
        value: otherText,
        isOther: true,
        otherText,
      },
    }));
  };

  // Handle field blur
  const handleBlur = (questionId: string) => {
    setTouchedFields(prev => new Set(prev).add(questionId));

    const question = prompt.questions.find(q => q.id === questionId);
    if (question) {
      const error = validateAnswer(question, answers[questionId]);
      if (error) {
        setErrors(prev => ({ ...prev, [questionId]: error }));
      }
    }
  };

  // Expose submit and validation methods to parent via ref
  useImperativeHandle(ref, () => ({
    handleSubmit: () => {
      // Mark all fields as touched
      const allQuestionIds = new Set(prompt.questions.map(q => q.id));
      setTouchedFields(allQuestionIds);

      // Validate all
      const validation = validateAll();

      if (!validation.isValid) {
        setErrors(validation.errors);
        // Scroll to first error
        const firstErrorId = Object.keys(validation.errors)[0];
        const element = document.getElementById(`inline-question-${firstErrorId}`);
        element?.scrollIntoView({ behavior: 'smooth', block: 'center' });
        return;
      }

      onSubmit(answers);
    },
    isValid: () => {
      return validateAll().isValid;
    },
  }));

  // Render different question types
  const renderQuestion = (question: InteractiveQuestion) => {
    const answer = answers[question.id];
    const error = touchedFields.has(question.id) ? errors[question.id] : null;
    const hasError = Boolean(error);

    switch (question.type) {
      case 'text':
        return (
          <div
            key={question.id}
            id={`inline-question-${question.id}`}
            className="inline-prompt-question"
          >
            <label htmlFor={`inline-input-${question.id}`} className="inline-prompt-label">
              {question.label}
              {question.required && <span className="inline-required-star">*</span>}
            </label>
            <input
              id={`inline-input-${question.id}`}
              type="text"
              className={`inline-prompt-input ${hasError ? 'error' : ''}`}
              placeholder={question.placeholder}
              value={(answer?.value as string) || ''}
              onChange={e => handleAnswerChange(question.id, e.target.value)}
              onBlur={() => handleBlur(question.id)}
              disabled={isSubmitting}
              maxLength={question.validation?.max_length}
            />
            {hasError && (
              <div className="inline-prompt-error">
                <AlertCircle size={14} />
                <span>{error}</span>
              </div>
            )}
            {question.validation?.max_length && (
              <div className="inline-prompt-hint">
                {((answer?.value as string) || '').length} / {question.validation.max_length}
              </div>
            )}
          </div>
        );

      case 'number':
        return (
          <div
            key={question.id}
            id={`inline-question-${question.id}`}
            className="inline-prompt-question"
          >
            <label htmlFor={`inline-input-${question.id}`} className="inline-prompt-label">
              {question.label}
              {question.required && <span className="inline-required-star">*</span>}
            </label>
            <input
              id={`inline-input-${question.id}`}
              type="number"
              className={`inline-prompt-input ${hasError ? 'error' : ''}`}
              placeholder={question.placeholder}
              value={(answer?.value as number) ?? ''}
              onChange={e => handleAnswerChange(question.id, e.target.valueAsNumber || 0)}
              onBlur={() => handleBlur(question.id)}
              disabled={isSubmitting}
              min={question.validation?.min}
              max={question.validation?.max}
            />
            {hasError && (
              <div className="inline-prompt-error">
                <AlertCircle size={14} />
                <span>{error}</span>
              </div>
            )}
          </div>
        );

      case 'checkbox':
        return (
          <div
            key={question.id}
            id={`inline-question-${question.id}`}
            className="inline-prompt-question"
          >
            <label className="inline-prompt-checkbox-label">
              <input
                type="checkbox"
                className="inline-prompt-checkbox"
                checked={(answer?.value as boolean) || false}
                onChange={e => handleAnswerChange(question.id, e.target.checked)}
                onBlur={() => handleBlur(question.id)}
                disabled={isSubmitting}
              />
              <span>
                {question.label}
                {question.required && <span className="inline-required-star">*</span>}
              </span>
            </label>
            {hasError && (
              <div className="inline-prompt-error">
                <AlertCircle size={14} />
                <span>{error}</span>
              </div>
            )}
          </div>
        );

      case 'select':
        const selectedValue = answer?.value as string;
        const isOtherSelected = answer?.isOther || false;

        return (
          <div
            key={question.id}
            id={`inline-question-${question.id}`}
            className="inline-prompt-question"
          >
            <label className="inline-prompt-label">
              {question.label}
              {question.required && <span className="inline-required-star">*</span>}
            </label>
            <div className="inline-prompt-radio-group">
              {question.options?.map(option => (
                <label key={option} className="inline-prompt-radio-label">
                  <input
                    type="radio"
                    name={`inline-radio-${question.id}`}
                    className="inline-prompt-radio"
                    value={option}
                    checked={selectedValue === option && !isOtherSelected}
                    onChange={() => handleAnswerChange(question.id, option, false)}
                    onBlur={() => handleBlur(question.id)}
                    disabled={isSubmitting}
                  />
                  <span>{option}</span>
                </label>
              ))}
              {question.allow_other && (
                <div className="inline-prompt-other-wrapper">
                  <label className="inline-prompt-radio-label">
                    <input
                      type="radio"
                      name={`inline-radio-${question.id}`}
                      className="inline-prompt-radio"
                      checked={isOtherSelected}
                      onChange={() => {
                        handleAnswerChange(question.id, answer?.otherText || '', true);
                      }}
                      disabled={isSubmitting}
                    />
                    <span>Other</span>
                  </label>
                  {isOtherSelected && (
                    <input
                      type="text"
                      className={`inline-prompt-input inline-prompt-other-input ${hasError ? 'error' : ''}`}
                      placeholder="Please specify..."
                      value={answer?.otherText || ''}
                      onChange={e => handleOtherTextChange(question.id, e.target.value)}
                      onBlur={() => handleBlur(question.id)}
                      disabled={isSubmitting}
                      autoFocus
                    />
                  )}
                </div>
              )}
            </div>
            {hasError && (
              <div className="inline-prompt-error">
                <AlertCircle size={14} />
                <span>{error}</span>
              </div>
            )}
          </div>
        );

      case 'multi-select':
        const selectedValues = (answer?.value as string[]) || [];
        const isOtherChecked = answer?.isOther || false;

        return (
          <div
            key={question.id}
            id={`inline-question-${question.id}`}
            className="inline-prompt-question"
          >
            <label className="inline-prompt-label">
              {question.label}
              {question.required && <span className="inline-required-star">*</span>}
            </label>
            <div className="inline-prompt-checkbox-group">
              {question.options?.map(option => (
                <label key={option} className="inline-prompt-checkbox-label">
                  <input
                    type="checkbox"
                    className="inline-prompt-checkbox"
                    value={option}
                    checked={selectedValues.includes(option)}
                    onChange={e => {
                      const newValues = e.target.checked
                        ? [...selectedValues, option]
                        : selectedValues.filter(v => v !== option);
                      handleAnswerChange(question.id, newValues, false);
                    }}
                    onBlur={() => handleBlur(question.id)}
                    disabled={isSubmitting}
                  />
                  <span>{option}</span>
                </label>
              ))}
              {question.allow_other && (
                <div className="inline-prompt-other-wrapper">
                  <label className="inline-prompt-checkbox-label">
                    <input
                      type="checkbox"
                      className="inline-prompt-checkbox"
                      checked={isOtherChecked}
                      onChange={e => {
                        if (e.target.checked) {
                          handleAnswerChange(question.id, [...selectedValues], true);
                        } else {
                          const newValues = selectedValues.filter(v => v !== answer?.otherText);
                          handleAnswerChange(question.id, newValues, false);
                        }
                      }}
                      disabled={isSubmitting}
                    />
                    <span>Other</span>
                  </label>
                  {isOtherChecked && (
                    <input
                      type="text"
                      className={`inline-prompt-input inline-prompt-other-input ${hasError ? 'error' : ''}`}
                      placeholder="Please specify..."
                      value={answer?.otherText || ''}
                      onChange={e => {
                        const otherText = e.target.value;
                        const newValues = selectedValues.filter(v => v !== answer?.otherText);
                        if (otherText) {
                          newValues.push(otherText);
                        }
                        setAnswers(prev => ({
                          ...prev,
                          [question.id]: {
                            questionId: question.id,
                            value: newValues,
                            isOther: true,
                            otherText,
                          },
                        }));
                      }}
                      onBlur={() => handleBlur(question.id)}
                      disabled={isSubmitting}
                      autoFocus
                    />
                  )}
                </div>
              )}
            </div>
            {hasError && (
              <div className="inline-prompt-error">
                <AlertCircle size={14} />
                <span>{error}</span>
              </div>
            )}
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <motion.div
      ref={cardRef}
      className="inline-prompt-card"
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -10 }}
      transition={{ duration: 0.3 }}
    >
      <div className="inline-prompt-header">
        <HelpCircle size={24} className="inline-prompt-icon" />
        <h3 className="inline-prompt-title">{prompt.title}</h3>
      </div>

      {prompt.description && <p className="inline-prompt-description">{prompt.description}</p>}

      <div className="inline-prompt-questions">{prompt.questions.map(renderQuestion)}</div>

      <div className="inline-prompt-footer">
        <p className="inline-prompt-hint-text">
          {prompt.allowSkip
            ? 'Answer the questions below or skip to continue.'
            : 'Please answer all required questions.'}
        </p>
      </div>
    </motion.div>
  );
});

InteractivePromptMessage.displayName = 'InteractivePromptMessage';
