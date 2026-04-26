import { entryFunction } from "./entry";

entryFunction();

export default function defaultHandler() {
  return "default";
}

export function entryFunction() {
  return "entry";
}