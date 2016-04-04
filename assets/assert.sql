create schema assert;

create table assert.test_table(
	id int
);

create function assert.test_func() returns void as $$
	begin 

	end;
$$ language plpgsql;

grant usage on schema assert to public;
grant execute on all functions in schema assert to public;