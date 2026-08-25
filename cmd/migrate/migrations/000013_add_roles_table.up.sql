CREATE TABLE IF NOT EXISTS roles(
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    level int NOT NULL DEFAULT 0,
    description TEXT
);
INSERT INTO roles (name, level, description)
VALUES (
        'user',
        1,
        'user can create posts and comments'
    );
INSERT INTO roles (name, level, description)
VALUES (
        'moderator',
        2,
        'moderator can update other user posts and comments'
    );
INSERT INTO roles (name, level, description)
VALUES (
        'admin',
        3,
        'admin can update and delete other user posts and comments'
    );