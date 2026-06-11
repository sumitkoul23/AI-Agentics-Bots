/**
 * Store — durable JSON-file persistence for deployments.
 * Survives restarts; on a host with a mounted disk it survives redeploys.
 */
import { readFileSync, writeFileSync, mkdirSync, existsSync } from "node:fs";
import { dirname } from "node:path";

export class Store {
  constructor(file) {
    this.file = file;
    this._items = [];
    mkdirSync(dirname(file), { recursive: true });
    if (existsSync(file)) {
      try { this._items = JSON.parse(readFileSync(file, "utf8")) || []; }
      catch { this._items = []; }
    }
  }

  _save() {
    writeFileSync(this.file, JSON.stringify(this._items, null, 2));
  }

  add(record) {
    this._items.unshift(record);
    this._save();
    return record;
  }

  all() {
    return this._items;
  }

  get(id) {
    return this._items.find((d) => d.id === id) || null;
  }

  stats() {
    return {
      deployments: this._items.length,
      validators: this._items.reduce((s, d) => s + (d.config?.validators || 0), 0),
      supply_sky: this._items.reduce((s, d) => s + (d.config?.total_supply_sky || 0), 0),
      framework: "Cosmos SDK + CometBFT",
    };
  }
}
