CREATE FUNCTION app.widget_count()
RETURNS bigint
LANGUAGE sql
AS $$
  SELECT count(*) FROM app.widgets;
$$;
