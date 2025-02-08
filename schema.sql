CREATE TABLE book(
    id SERIAL PRIMARY KEY,
    title varchar(20),
    year int
);

CREATE TABLE author(
    id SERIAL PRIMARY KEY,
    name varchar(40),
    biography text
);

CREATE TABLE attribution(
    author_id int REFERENCES author(id),
    book_id int REFERENCES book(id),
    PRIMARY KEY(author_id, book_id)
);

INSERT INTO author(name, biography) VALUES
('George Orwell', 'British writer known for 1984 and Animal Farm')
;

INSERT INTO book(title, year) VALUES
('1984', 1945)
,('Animal Farm', 1949)
;

INSERT INTO attribution VALUES
(1, 1)
,(1, 2)
;
