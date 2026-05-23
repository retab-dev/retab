#nullable enable

namespace Retab
{
    /// <summary>
    /// Generated service accessors for the <see cref="Retab"/> client.
    /// </summary>
    public partial class Retab
    {
        private SchemasService? schemas;

        /// <summary>Gets the <see cref="SchemasService"/> for schemas API operations.</summary>
        public virtual SchemasService Schemas => this.schemas ??= new SchemasService(this);

        private ExtractionsService? extractions;

        /// <summary>Gets the <see cref="ExtractionsService"/> for extractions API operations.</summary>
        public virtual ExtractionsService Extractions => this.extractions ??= new ExtractionsService(this);

        private ClassificationsService? classifications;

        /// <summary>Gets the <see cref="ClassificationsService"/> for classifications API operations.</summary>
        public virtual ClassificationsService Classifications => this.classifications ??= new ClassificationsService(this);

        private ParsesService? parses;

        /// <summary>Gets the <see cref="ParsesService"/> for parses API operations.</summary>
        public virtual ParsesService Parses => this.parses ??= new ParsesService(this);

        private PartitionsService? partitions;

        /// <summary>Gets the <see cref="PartitionsService"/> for partitions API operations.</summary>
        public virtual PartitionsService Partitions => this.partitions ??= new PartitionsService(this);

        private SplitsService? splits;

        /// <summary>Gets the <see cref="SplitsService"/> for splits API operations.</summary>
        public virtual SplitsService Splits => this.splits ??= new SplitsService(this);

        private FilesService? files;

        /// <summary>Gets the <see cref="FilesService"/> for files API operations.</summary>
        public virtual FilesService Files => this.files ??= new FilesService(this);

        private WorkflowRunsService? workflowRuns;

        /// <summary>Gets the <see cref="WorkflowRunsService"/> for workflow runs API operations.</summary>
        public virtual WorkflowRunsService WorkflowRuns => this.workflowRuns ??= new WorkflowRunsService(this);

        private WorkflowStepsService? workflowSteps;

        /// <summary>Gets the <see cref="WorkflowStepsService"/> for workflow steps API operations.</summary>
        public virtual WorkflowStepsService WorkflowSteps => this.workflowSteps ??= new WorkflowStepsService(this);

        private WorkflowReviewsService? workflowReviews;

        /// <summary>Gets the <see cref="WorkflowReviewsService"/> for workflow reviews API operations.</summary>
        public virtual WorkflowReviewsService WorkflowReviews => this.workflowReviews ??= new WorkflowReviewsService(this);

        private WorkflowReviewVersionsService? workflowReviewVersions;

        /// <summary>Gets the <see cref="WorkflowReviewVersionsService"/> for workflow review versions API operations.</summary>
        public virtual WorkflowReviewVersionsService WorkflowReviewVersions => this.workflowReviewVersions ??= new WorkflowReviewVersionsService(this);

        private WorkflowArtifactsService? workflowArtifacts;

        /// <summary>Gets the <see cref="WorkflowArtifactsService"/> for workflow artifacts API operations.</summary>
        public virtual WorkflowArtifactsService WorkflowArtifacts => this.workflowArtifacts ??= new WorkflowArtifactsService(this);

        private WorkflowTestRunsService? workflowTestRuns;

        /// <summary>Gets the <see cref="WorkflowTestRunsService"/> for workflow test runs API operations.</summary>
        public virtual WorkflowTestRunsService WorkflowTestRuns => this.workflowTestRuns ??= new WorkflowTestRunsService(this);

        private WorkflowTestRunResultsService? workflowTestRunResults;

        /// <summary>Gets the <see cref="WorkflowTestRunResultsService"/> for workflow test run results API operations.</summary>
        public virtual WorkflowTestRunResultsService WorkflowTestRunResults => this.workflowTestRunResults ??= new WorkflowTestRunResultsService(this);

        private WorkflowBlocksService? workflowBlocks;

        /// <summary>Gets the <see cref="WorkflowBlocksService"/> for workflow blocks API operations.</summary>
        public virtual WorkflowBlocksService WorkflowBlocks => this.workflowBlocks ??= new WorkflowBlocksService(this);

        private WorkflowEdgesService? workflowEdges;

        /// <summary>Gets the <see cref="WorkflowEdgesService"/> for workflow edges API operations.</summary>
        public virtual WorkflowEdgesService WorkflowEdges => this.workflowEdges ??= new WorkflowEdgesService(this);

        private WorkflowTestsService? workflowTests;

        /// <summary>Gets the <see cref="WorkflowTestsService"/> for workflow tests API operations.</summary>
        public virtual WorkflowTestsService WorkflowTests => this.workflowTests ??= new WorkflowTestsService(this);

        private ExperimentRunsService? experimentRuns;

        /// <summary>Gets the <see cref="ExperimentRunsService"/> for experiment runs API operations.</summary>
        public virtual ExperimentRunsService ExperimentRuns => this.experimentRuns ??= new ExperimentRunsService(this);

        private ExperimentRunResultsService? experimentRunResults;

        /// <summary>Gets the <see cref="ExperimentRunResultsService"/> for experiment run results API operations.</summary>
        public virtual ExperimentRunResultsService ExperimentRunResults => this.experimentRunResults ??= new ExperimentRunResultsService(this);

        private ExperimentRunMetricsService? experimentRunMetrics;

        /// <summary>Gets the <see cref="ExperimentRunMetricsService"/> for experiment run metrics API operations.</summary>
        public virtual ExperimentRunMetricsService ExperimentRunMetrics => this.experimentRunMetrics ??= new ExperimentRunMetricsService(this);

        private WorkflowExperimentsService? workflowExperiments;

        /// <summary>Gets the <see cref="WorkflowExperimentsService"/> for workflow experiments API operations.</summary>
        public virtual WorkflowExperimentsService WorkflowExperiments => this.workflowExperiments ??= new WorkflowExperimentsService(this);

        private WorkflowBlockExecutionsService? workflowBlockExecutions;

        /// <summary>Gets the <see cref="WorkflowBlockExecutionsService"/> for workflow block executions API operations.</summary>
        public virtual WorkflowBlockExecutionsService WorkflowBlockExecutions => this.workflowBlockExecutions ??= new WorkflowBlockExecutionsService(this);

        private WorkflowsService? workflows;

        /// <summary>Gets the <see cref="WorkflowsService"/> for workflows API operations.</summary>
        public virtual WorkflowsService Workflows => this.workflows ??= new WorkflowsService(this);

        private WorkflowSpecsService? workflowSpecs;

        /// <summary>Gets the <see cref="WorkflowSpecsService"/> for workflow specs API operations.</summary>
        public virtual WorkflowSpecsService WorkflowSpecs => this.workflowSpecs ??= new WorkflowSpecsService(this);

        private EditTemplatesService? editTemplates;

        /// <summary>Gets the <see cref="EditTemplatesService"/> for edit templates API operations.</summary>
        public virtual EditTemplatesService EditTemplates => this.editTemplates ??= new EditTemplatesService(this);

        private EditsService? edits;

        /// <summary>Gets the <see cref="EditsService"/> for edits API operations.</summary>
        public virtual EditsService Edits => this.edits ??= new EditsService(this);

        private JobsService? jobs;

        /// <summary>Gets the <see cref="JobsService"/> for jobs API operations.</summary>
        public virtual JobsService Jobs => this.jobs ??= new JobsService(this);

    }
}
