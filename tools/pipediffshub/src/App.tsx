import { ReviewUI } from './diffshub-copied/ReviewUI';
import { useDiffMeta } from './diffMeta';
import { useHeartbeat } from './useHeartbeat';

export function App() {
  useHeartbeat();
  const meta = useDiffMeta();

  return (
    <div className="flex h-dvh flex-col">
      <ReviewUI path="/piped-diff" meta={meta} />
    </div>
  );
}
