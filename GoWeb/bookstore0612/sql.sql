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