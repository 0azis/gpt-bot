CREATE TABLE users (
    id bigint not null,
    avatar varchar(255) not null,
    balance int default 150 not null,
    language_code varchar(255) not null,
    referral_code varchar(5),
    referred_by varchar(5),
    created_at timestamp default current_timestamp,
    primary key (id)
);

CREATE TABLE chats (
    id smallint not null auto_increment,
    user_id bigint not null,
    title varchar(255),
    model enum (
        'o1-preview',
        'gpt-4o',
        'o1-mini',
        'gpt-4o-mini',
        'dall-e-3',
        'runware'
    ) not null,
    type enum ('text', 'image') not null,
    foreign key (user_id) references users (id) on delete cascade,
    primary key (id)
);

CREATE TABLE messages (
    id smallint not null auto_increment,
    chat_id smallint not null,
    content text not null,
    role enum ('user', 'assistant') not null,
    type enum ('text', 'image') not null,
    created_at timestamp default current_timestamp,
    foreign key (chat_id) references chats (id) on delete cascade,
    primary key (id)
);

CREATE TABLE bonuses (
    id smallint not null auto_increment,
    name varchar(255) default (''),
    channel_id bigint signed default 0,
    max_users smallint default 0,
    award int default 0,
    link varchar(255) default (''),
    is_check bool default 0,
    created_at date default (current_date()),
    primary key (id)
);

CREATE TABLE user_bonuses (
    bonus_id smallint not null,
    user_id bigint not null,
    awarded bool default false not null,
    awarded_at date null,
    foreign key (bonus_id) references bonuses (id) on delete cascade,
    foreign key (user_id) references users (id) on delete cascade,
    primary key (user_id, bonus_id)
);

CREATE TABLE subscriptions_info (
    name enum ('standard', 'advanced', 'ultimate') not null,
    diamonds smallint not null,
    primary key (name)
);

CREATE TABLE subscriptions (
    user_id bigint not null,
    name enum ('standard', 'advanced', 'ultimate') default 'standard' not null,
    start date default (current_date()),
    end date default null,
    foreign key (user_id) references users (id) on delete cascade,
    primary key (user_id)
);

CREATE TABLE limits (
    user_id bigint not null,
    o1_preview int not null,
    gpt_4o int not null,
    o1_mini int not null,
    gpt_4o_mini int not null,
    runware int not null,
    dall_e_3 int not null,
    foreign key (user_id) references users (id) on delete cascade,
    primary key (user_id)
);

CREATE TABLE referrals (
    id smallint not null auto_increment,
    name varchar(255) not null,
    code varchar(255) not null,
    primary key (id)
);

CREATE TABLE user_referrals (
    referral_id smallint not null,
    user_id bigint not null,
    created_at date,
    foreign key (referral_id) references referrals (id) on delete cascade,
    foreign key (user_id) references users (id) on delete no action,
    primary key (referral_id, user_id)
);

CREATE TABLE stats (user_id bigint not null, clicked_at date);

CREATE TABLE admins (user_id bigint not null);

INSERT INTO
    subscriptions_info (name, diamonds)
values
    ('standard', 150);

INSERT INTO
    subscriptions_info (name, diamonds)
values
    ('advanced', 1500);

INSERT INTO
    subscriptions_info (name, diamonds)
values
    ('ultimate', 4000);

INSERT INTO
    admins (user_id)
values
    (7373317122);

INSERT INTO
    admins (user_id)
values
    (992956951);
