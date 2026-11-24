CREATE TABLE IF NOT EXISTS posts(
    id bigserial PRIMARY KEY,
    content text NOT NULL,
    title varchar(255) NOT NULL,
    user_id bigint NOT NULL REFERENCES users(id),
    tags varchar(100)[],
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);
