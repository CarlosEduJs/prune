import type { User } from '../types/user';

export const Dashboard = ({ user }: { user: User }) => {
    return (
        <div className="dashboard">
            <h1>Welcome, {user.name}</h1>
            <p>Email: {user.email}</p>
        </div>
    );
};

export const UnusedComponent = () => { // not used
    return <div>I'm dead</div>;
};
