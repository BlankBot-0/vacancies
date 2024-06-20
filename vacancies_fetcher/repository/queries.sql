-- name: AddVacancies :exec
insert into vacancies (key_word, href, title)
values (@key_word::text, unnest(@hrefs::text[]), unnest(@titles::text[]));

-- name: GetVacanciesByKeyWord :many
select href, title, max(id) from vacancies
where key_word = @key_word::text
limit @limit_val::int offset @offset_val::int;