create table cart_items(
                           id int primary key auto_increment,
                           count int not null ,
                           amount double(11,2) not null,
                           book_id int not null,
                           cart_id varchar(100) not null,
                           foreign key(book_id) references books(id),
                           foreign key(cart_id) references carts(id)
);

create table carts(
                      id varchar(100) primary key ,
                      total_count int not null ,
                      tatal_amount double(11,2) not null ,
                      user_id int not null,
                      foreign key(user_id) references users(id)
);
create table orders(
                       id varchar(100) primary key ,
                       create_time DATETIME NOT NULL ,
                       total_count int not null ,
                       total_amount double(11,2),
                       state int not null,
                       user_id int,
                       foreign key(user_id) references users(id)
);
create table order_items(
                            id int primary key  auto_increment,
                            count int not null ,
                            amount double(11,2) not null,
                            title varchar(100) not null,
                            author varchar(100) not null ,
                            price varchar(100) not null,
                            img_path varchar(100) not null ,
                            order_id varchar(100) not null,
                            foreign key(order_id) references orders(id)
);