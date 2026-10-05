type SnakeCase<S extends string> = S extends `${infer Head}${infer Tail}`
  ? Head extends Lowercase<Head>
    ? `${Head}${SnakeCase<Tail>}`
    : `_${Lowercase<Head>}${SnakeCase<Tail>}`
  : S;

export type ObservabilityResponse<T> = {
  [Key in keyof T as Key extends string ? SnakeCase<Key> : Key]: T[Key];
};

export function normalizeObservability<T>(input: ObservabilityResponse<T>): T {
  return Object.fromEntries(
    Object.entries(input as Record<string, unknown>).map(([key, value]) => [
      key.replace(/_([a-z])/g, (_, letter: string) => letter.toUpperCase()),
      value,
    ]),
  ) as T;
}

export function normalizeObservabilityList<T>(
  input: ObservabilityResponse<T>[] | null,
): T[] {
  return (input ?? []).map((item) => normalizeObservability<T>(item));
}
