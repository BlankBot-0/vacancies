-- +goose Up
-- +goose StatementBegin
create table vacancies (
    id bigserial primary key,
    href text not null,
    title text not null,
    key_word text not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table vacancies;
-- +goose StatementEnd
