export function parseSearchQuery(input: string): string[] {
  const terms: string[] = []
  let current = ''
  let quoted = false
  for (const character of input.trim()) {
    if (character === '"') {
      quoted = !quoted
      continue
    }
    if (/\s/.test(character) && !quoted) {
      if (current) terms.push(current)
      current = ''
      continue
    }
    current += character
  }
  if (quoted) throw new Error('Close the quoted search value.')
  if (current) terms.push(current)
  return terms
}
