import type { User, UserRole } from './types/user';
import { Database } from './services/db';
import { Dashboard } from './components/Dashboard';

async function bootstrap() {
    const user: User = {
        id: '1',
        name: 'Carlos',
        email: 'carlos@example.com'
    };
    
    const role: UserRole = 'admin';
    console.log(`Setting up for ${role}`);

    const db = new Database<User>();
    await db.save(user);

    const ui = Dashboard({ user });
    console.log('UI Rendered:', ui);
}

bootstrap().catch(console.error);
