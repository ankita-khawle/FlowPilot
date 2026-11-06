export type SummaryTone =
  | 'professional'
  | 'casual'
  | 'technical'
  | 'executive';

export type SummaryLength =
  | 'short'
  | 'medium'
  | 'detailed';

export interface SummaryResponse {
  summary: string;
  extracted_data?: Record<string, unknown>;
}