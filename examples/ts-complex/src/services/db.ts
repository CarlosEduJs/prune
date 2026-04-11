export interface Identifiable {
    id: string;
}

export class Database<T extends Identifiable> {
    private items: T[] = [];

    async save(item: T): Promise<void> {
        this.items.push(item);
    }

    async findById(id: string): Promise<T | undefined> {
        return this.items.find(i => i.id === id);
    }

    async clear(): Promise<void> { // not used
        this.items = [];
    }
}

export function unusedGeneric<T>(arg: T): T { // not used
    return arg;
}
