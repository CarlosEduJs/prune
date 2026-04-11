export function unusedExport() {
  return 42
}

export const unusedVar = "dead"

function unusedLocal() {
  return "local"
}

export default () => {
  return "default"
}
