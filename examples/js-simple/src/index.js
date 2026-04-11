import { used, usedWrapper, value } from "./used"
import { reexportedValue } from "./reexports"
import "./side-effect"

export function main() {
  return used() + usedWrapper() + value + reexportedValue
}

main()
console.log(flags.EXPERIMENTAL)
eval("console.log('dynamic')")

export default function run() {
  return main()
}

run()
