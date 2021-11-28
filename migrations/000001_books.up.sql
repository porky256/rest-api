create table if not exists genres(
                       id serial not null primary key,
                       name varchar(100) not null

);

insert into genres values (1,'Adventure');
insert into genres values (2,'Classics');
insert into genres values (3,'Fantasy');

create table if not exists books(
                      id serial not null primary key,
                      name varchar(100) not null unique,
                      price real not null,
                      genre int not null references genres(id),
                      amount int not null
);
