create table auth
(
    user_id    varchar(255) not null,
    method     varchar(255) not null,
    social_id  varchar(255) not null,
    created_at timestamp    not null,
    updated_at timestamp    not null
);

alter table auth
    owner to postgres;

create unique index auth_social_id_method_uindex
    on auth (social_id, method);

create table image
(
    id           varchar(255)              not null,
    storage_type varchar(255)              not null,
    file_id      text                      not null,
    width        integer                   not null,
    height       integer                   not null,
    hash         varchar(255),
    created_at   timestamp                 not null,
    sizes        jsonb default '[]'::jsonb not null
);

alter table image
    owner to postgres;

create unique index image_id_uindex
    on image (id);

CREATE TYPE price AS (
    number NUMERIC,
    currency_code TEXT
);

create table product
(
    id          varchar(255) not null,
    title       varchar(255) not null,
    image_id    varchar(255),
    description text,
    url         varchar(255),
    created_at  timestamp    not null,
    updated_at  timestamp    not null,
    price       price
);

alter table product
    owner to postgres;

create unique index product_id_uindex
    on product (id);

create table subscribe
(
    user_id     varchar(255) not null,
    wishlist_id varchar(255) not null,
    created_at  timestamp    not null
);

alter table subscribe
    owner to postgres;

create index subscribe_user_id_index
    on subscribe (user_id);

create table users
(
    id                varchar(255)                                not null,
    fullname          varchar(255)                                not null,
    created_at        timestamp                                   not null,
    lang              varchar(10) default 'en'::character varying not null,
    notify_type       varchar(255),
    notify_channel_id varchar(255)
);

alter table users
    owner to postgres;

create unique index users_id_uindex
    on users (id);

create table wishlist
(
    id          varchar(255) not null,
    user_id     varchar(255) not null,
    is_default  boolean      not null,
    title       varchar(255) not null,
    image_id    varchar(255),
    description text         not null,
    is_archived boolean      not null,
    created_at  timestamp    not null,
    updated_at  timestamp    not null
);

alter table wishlist
    owner to postgres;

create unique index wishlist_id_uindex
    on wishlist (id);

create index wishlist_user_id_index
    on wishlist (user_id);

create table wishlist_item
(
    wishlist_id          varchar(255) not null,
    product_id           varchar(255) not null,
    is_booking_available boolean      not null,
    is_booked_by         varchar(255),
    created_at           timestamp    not null,
    updated_at           timestamp    not null
);

alter table wishlist_item
    owner to postgres;

create unique index wishlist_item_wishlist_id_product_id_uindex
    on wishlist_item (wishlist_id, product_id);

create table file
(
    id         varchar(255) not null,
    content    bytea        not null,
    created_at timestamp    not null
);

alter table file
    owner to postgres;

create unique index file_id_uindex
    on file (id);

-- default image
INSERT INTO image
    (id, storage_type, file_id, width, height, hash, created_at, sizes)
VALUES
    ('0fc13627-7e95-4bde-ac63-e962969b921a','remote_link','https://wishlist.super-app.studio/api/images/dGVsZWdyYW1fYm90OkFnQUNBZ0lBQXhrREFBT1NaVVBMeU9wcnZ2U3VfMnFCU3A4SUl2Qzg2ZVVBQWpqVk1Sc1pPaUJLZEktZk9fWnBfS1FCQUFNQ0FBTnRBQU16QkE=',1000,1000,'a:f0f8fcfcfc7c0000;d:00207054504c4002;p:d6c929b6a49c9669',NOW(),'[]');
