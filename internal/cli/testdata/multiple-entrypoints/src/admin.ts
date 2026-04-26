import { shared } from "./main";

export function admin() {
  return "admin: " + shared();
}