CREATE TABLE users(
    id bigint not null,
    subscription enum('standard', 'advanced', 'ultimate') default 'standard' not null,
    avatar varchar(255) not null,
    balance int default 150 not null,
    referral_code varchar(255),
    referred_by varchar(255),
    primary key(id)
);

CREATE TABLE chats (
    id smallint not null auto_increment,
    user_id bigint not null,
    title varchar(255),
    model varchar(255) not null,
    type enum('text', 'image') not null,
    foreign key (user_id) references users (id) on delete cascade,
    primary key(id)
);

CREATE TABLE messages(
    id smallint not null auto_increment,
    chat_id smallint not null,
    content text not null,
    role enum('user', 'assistant') not null,
    foreign key (chat_id) references chats (id) on delete cascade,
    primary key(id)
);

CREATE TABLE bonuses(
    id smallint not null auto_increment,
    award tinyint not null,
    bonus_type varchar(255) not null,
    primary key(id)
);

CREATE TABLE subscriptions(
    name enum('standard', 'advanced', 'ultimate') not null,
    diamonds smallint not null,
    primary key(name)
);

INSERT INTO bonuses (award, bonus_type) values (10, 'referral');
INSERT INTO subscriptions(name, diamonds) values ('standard', 150);
INSERT INTO subscriptions(name, diamonds) values ('advanced', 1500);
INSERT INTO subscriptions(name, diamonds) values ('ultimate', 3000);
