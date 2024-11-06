CREATE TABLE users(
    id int not null,
    subscription bool default false not null,
    requests tinyint default 10 not null,
    primary key(id)
);

-- CREATE TABLE refs(
--     id smallint not null,
--     user_id smallint not null
-- );

-- CREATE TABLE chat(
--     id smallint not null,
--     user_id smallint not null,

--     primary key(id)
-- );
