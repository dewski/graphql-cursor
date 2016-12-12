CREATE TABLE graphql_records (
  id serial primary key
);

INSERT INTO graphql_records (id) VALUES
	(1),
	(2),
	(3),
	(4),
	(5),
	(6),
	(7),
	(8),
	(9),
	(10),
	(11),
	(12),
	(13),
	(14),
	(15),
	(16),
	(17),
	(18),
	(19),
	(20);

alter sequence graphql_records_id_seq restart with 100;
