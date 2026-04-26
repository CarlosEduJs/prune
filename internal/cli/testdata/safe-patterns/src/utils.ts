export function useConsole() {
  console.log("hello");
  return Math.random();
}

export function useJson() {
  return JSON.stringify({ a: 1 });
}

export function useArray() {
  return Array.isArray([]);
}

export function safeFunc() {
  return process.env.NODE_ENV;
}