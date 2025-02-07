CREATE TABLE book(
    id int PRIMARY KEY,
    title char(20),
    year int
);

CREATE TABLE author(
    id int PRIMARY KEY,
    name char(40),
    biography text
);

CREATE TABLE attribution(
    author_id int REFERENCES author(id),
    book_id int REFERENCES book(id),
    PRIMARY KEY(author_id, book_id)
);

INSERT INTO author VALUES
(1, 'George Orwell', 'British writer known for 1984 and Animal Farm')
;

INSERT INTO book VALUES
(1, '1984', 1945)
,(2, 'Animal Farm', 1949)
;

INSERT INTO attribution VALUES
(1, 1)
,(1, 2)
;
