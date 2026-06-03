#nullable enable

namespace Retab
{
    /// <summary>
    /// Generated service accessors for the <see cref="Retab"/> client.
    /// </summary>
    public partial class Retab
    {
        private ClassificationsService? classifications;

        /// <summary>Gets the <see cref="ClassificationsService"/> for classifications API operations.</summary>
        public virtual ClassificationsService Classifications => this.classifications ??= new ClassificationsService(this);

        private EditsService? edits;

        /// <summary>Gets the <see cref="EditsService"/> for edits API operations.</summary>
        public virtual EditsService Edits => this.edits ??= new EditsService(this);

        private ExtractionsService? extractions;

        /// <summary>Gets the <see cref="ExtractionsService"/> for extractions API operations.</summary>
        public virtual ExtractionsService Extractions => this.extractions ??= new ExtractionsService(this);

        private FilesService? files;

        /// <summary>Gets the <see cref="FilesService"/> for files API operations.</summary>
        public virtual FilesService Files => this.files ??= new FilesService(this);

        private ParsesService? parses;

        /// <summary>Gets the <see cref="ParsesService"/> for parses API operations.</summary>
        public virtual ParsesService Parses => this.parses ??= new ParsesService(this);

        private PartitionsService? partitions;

        /// <summary>Gets the <see cref="PartitionsService"/> for partitions API operations.</summary>
        public virtual PartitionsService Partitions => this.partitions ??= new PartitionsService(this);

        private SchemasService? schemas;

        /// <summary>Gets the <see cref="SchemasService"/> for schemas API operations.</summary>
        public virtual SchemasService Schemas => this.schemas ??= new SchemasService(this);

        private SecretsService? secrets;

        /// <summary>Gets the <see cref="SecretsService"/> for secrets API operations.</summary>
        public virtual SecretsService Secrets => this.secrets ??= new SecretsService(this);

        private SplitsService? splits;

        /// <summary>Gets the <see cref="SplitsService"/> for splits API operations.</summary>
        public virtual SplitsService Splits => this.splits ??= new SplitsService(this);

        private TablesService? tables;

        /// <summary>Gets the <see cref="TablesService"/> for tables API operations.</summary>
        public virtual TablesService Tables => this.tables ??= new TablesService(this);

        private WorkflowsService? workflows;

        /// <summary>Gets the <see cref="WorkflowsService"/> for workflows API operations.</summary>
        public virtual WorkflowsService Workflows => this.workflows ??= new WorkflowsService(this);

    }
}
