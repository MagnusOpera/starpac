CREATE TABLE app.widgets (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name app.widget_name,
  status app.widget_status NOT NULL DEFAULT 'new'
);
