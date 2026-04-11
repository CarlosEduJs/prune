export interface User {
    id: string;
    name: string;
    email: string;
}

export type UserRole = 'admin' | 'user' | 'guest';

export interface UnusedType { // not used
    foo: string;
}
