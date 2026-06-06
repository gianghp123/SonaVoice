type JsonObject = Record<string, unknown>

export function snakeToCamel<T = unknown>(obj: unknown): T {
  if (obj === null || obj === undefined) {
    return obj as T
  }
  if (Array.isArray(obj)) {
    return obj.map((item) => snakeToCamel(item)) as T
  }
  if (typeof obj !== "object") {
    return obj as T
  }

  const result: JsonObject = {}

  for (const [key, value] of Object.entries(obj)) {
    const camelKey = key.replace(/_([a-z])/g, (_, letter: string) =>
      letter.toUpperCase()
    )
    result[camelKey] = snakeToCamel(value)
  }

  return result as T
}

export function camelToSnake<T = unknown>(obj: unknown): T {
  if (obj === null || obj === undefined) {
    return obj as T
  }
  if (Array.isArray(obj)) {
    return obj.map((item) => camelToSnake(item)) as T
  }
  if (typeof obj !== "object") {
    return obj as T
  }

  const result: JsonObject = {}

  for (const [key, value] of Object.entries(obj)) {
    const snakeKey = key.replace(/[A-Z]/g, (letter) => `_${letter.toLowerCase()}`)
    result[snakeKey] = camelToSnake(value)
  }

  return result as T
}