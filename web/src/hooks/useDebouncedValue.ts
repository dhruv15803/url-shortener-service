import { useEffect, useState } from "react"

/**
 * Trails `value` by `delayMs`, resetting the timer on every change. Used to
 * keep a search box responsive to typing without firing a request per
 * keystroke.
 */
export function useDebouncedValue<T>(value: T, delayMs = 300): T {
  const [debounced, setDebounced] = useState(value)

  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delayMs)
    return () => clearTimeout(timer)
  }, [value, delayMs])

  return debounced
}
