CREATE TRIGGER widgets_created
AFTER INSERT ON widgets
BEGIN
  INSERT INTO widget_events(widget_id, event)
  VALUES (NEW.id, 'created');
END;
