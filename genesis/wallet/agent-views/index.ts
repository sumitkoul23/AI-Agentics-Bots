// Barrel export. Lets the Keplr fork and the DEX frontend both write
// `import { Registry, Tasks, Reputation, Streams } from "@/agent-views"`.
export { Registry, type AgentViewProps } from "./Registry";
export { Tasks } from "./Tasks";
export { Reputation } from "./Reputation";
export { Streams } from "./Streams";
export { fmtGen } from "./hooks";
