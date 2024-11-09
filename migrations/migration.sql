CREATE TABLE users(
    id bigint not null,
    subscription varchar(255) default 'standard' not null,
    requests tinyint default 10 not null,
    avatar varchar(255) not null,
    balance int default 0 not null,
    referral_code varchar(255),
    referred_by varchar(255),
    primary key(id)
);

CREATE TABLE chats (
    id smallint not null auto_increment,
    user_id bigint not null,
    title varchar(255),
    model varchar(255) not null,
    type enum('chat', 'image') not null,
    foreign key (user_id) references users (id) on delete cascade,
    primary key(id)
);

CREATE TABLE messages(
    id smallint not null auto_increment,
    chat_id smallint not null,
    content text not null,
    is_user bool default false not null,
    foreign key (chat_id) references chats (id) on delete cascade,
    primary key(id)
);

CREATE TABLE bonuses(
    id smallint not null auto_increment,
    award tinyint not null,
    bonus_type varchar(255) not null,
    primary key(id)
);

INSERT INTO bonuses (award, bonus_type) values (10, 'referral')
