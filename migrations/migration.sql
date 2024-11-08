CREATE TABLE users(
    id int not null,
    subscription varchar(255) default 'standart' not null,
    requests tinyint default 10 not null,
    avatar varchar(255) not null,
    balance int default 0 not null,
    referral_code varchar(255) not null,
    referred_by varchar(255) not null,
    primary key(id)
);

CREATE TABLE chats (
    id smallint not null auto_increment,
    user_id int not null,
    title varchar(255),
    model varchar(255) not null,
    type varchar(255) not null,
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
    user_id int not null,
    award tinyint not null,
    bonus_type varchar(255) not null,
    primary key(id)
);

-- CREATE TABLE referrals (
--     user_id int not null,
--     ref_user int not null,
--     foreign key (user_id) references users(id) on delete no action,
--     foreign key (ref_user) references users(id) on delete no action,
--     primary key(user_id, ref_user)
-- );
