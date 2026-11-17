
import { apiRequest, ApiError } from './apiClient';
import type {
  AnswerResponse,
  AskDocumentRequest,
  Citation,
  IngestResult,
} from './types';

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function isIngestResult(value: unknown): value is IngestResult {
  if (!isRecord(value)) return false;

  return (
    typeof value.document_id === 'string' &&
    value.document_id.length > 0 &&
    typeof value.file_name === 'string' &&
    typeof value.file_type === 'string' &&
    typeof value.status === 'string' &&
    typeof value.word_count === 'number' &&
    typeof value.character_count === 'number' &&
    typeof value.chunk_count === 'number'
  );
}

function isCitation(value: unknown): value is Citation {
  if (!isRecord(value)) return false;

  return (
    typeof value.source === 'string' &&
    (value.page === undefined || typeof value.page === 'number') &&
    (value.quote === undefined || typeof value.quote === 'string') &&
    (value.field === undefined || typeof value.field === 'string')
  );
}

function isAnswerResponse(value: unknown): value is AnswerResponse {
  if (!isRecord(value)) return false;

  return (
    typeof value.answer === 'string' &&
    Array.isArray(value.citations) &&
    value.citations.every(isCitation)
  );
}

export async function uploadDocument(
  file: File
): Promise<IngestResult> {
  const formData = new FormData();
  formData.append('file', file);

  const result = await apiRequest<unknown>(
    '/api/rag/documents',
    {
      method: 'POST',
      body: formData,
    }
  );

  if (!isIngestResult(result)) {
    throw new ApiError(
      'The server returned an unexpected document response.',
      200,
      'INVALID_RESPONSE'
    );
  }

  return result;
}

export async function askDocument(
  documentId: string,
  request: AskDocumentRequest
): Promise<AnswerResponse> {
  if (!documentId.trim()) {
    throw new Error('A valid document ID is required.');
  }

  if (!request.question.trim()) {
    throw new Error('Please enter a question.');
  }

  const result = await apiRequest<unknown>(
    `/api/rag/documents/${encodeURIComponent(documentId)}/ask`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        question: request.question.trim(),
        limit: request.limit ?? 5,
      }),
    }
  );

  if (!isAnswerResponse(result)) {
    throw new ApiError(
      'The server returned an unexpected answer response.',
      200,
      'INVALID_RESPONSE'
    );
  }

  return result;
}
