namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a declarative plan change.</summary>
    public class DeclarativePlanChange
    {
        public object? Before { get; set; }
        public object? After { get; set; }
        public object? BeforeSensitive { get; set; }
        public object? AfterSensitive { get; set; }
        public List<DeclarativePlanFieldChange>? FieldChanges { get; set; }
    }
}
