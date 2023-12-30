create table auth_users
(
    id   uuid not null
        primary key,
    username text,
    password text,
    role text
);

create table users
(
    id   uuid not null
        primary key,
    name text,
    surname text,
    phone text,
    address text,
    subscription boolean,
    subscription_symbols text[],
    user_id           uuid
        constraint fk_auth_user
            references "auth_users"
);