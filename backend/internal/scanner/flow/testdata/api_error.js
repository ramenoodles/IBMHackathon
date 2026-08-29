export function messageFromApiFailure(status, body) {
  if (status === 429) {
    const seconds = parseRetryAfter(body);
    if (Number.isFinite(seconds) && seconds > 0) {
      return `Rate limited. Retry in ${seconds}s`;
    }
  }
  if (status === 401 || status === 403) {
    return 'Unauthorized response';
  }
  if (status === 502 || status === 503 || status === 504) {
    return 'Gateway error';
  }
  return String(body);
}
