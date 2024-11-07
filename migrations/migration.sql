CREATE TABLE users(
    id int not null,
    subscription bool default false not null,
    requests tinyint default 10 not null,
    avatar varchar(255) not null,
    primary key(id)
);

CREATE TABLE chats (
    id smallint not null auto_increment,
    user_id int not null,
    title varchar(255) not null,
    primary key(id)
);

CREATE TABLE messages(
    id smallint not null auto_increment,
    chat_id smallint not null,
    content text not null,
    is_user bool default false not null,
    primary key(id)
);
