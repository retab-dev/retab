from pydantic import BaseModel

from ..._utils.benchmarking import EvalMetrics, SingleFileEval, compute_dict_difference
from .batch_annotation import AnnotationInputData, InferenceSettings

# This job will generate two datasets from the original dataset, one with the first annotation and one with the second annotation
# It will then evaluate the two datasets using the evaluation metrics and return an EvalMetrics object


class EvaluationInputData(BaseModel):
    original_dataset_id: str
    schema_id: str
    schema_data_id: str
    inference_settings_1: InferenceSettings
    inference_settings_2: InferenceSettings


# def evaluate_datasets(
#     original_dataset_id: str,
#     inference_settings_1: InferenceSettings,
#     inference_settings_2: InferenceSettings,
#     identity: Identity,
#     job_execution_id: str,
#     settings: Settings,
#     dashboard_db: AsyncIOMotorDatabase,
# ) -> EvalMetrics:
#     # Generate two datasets from the original dataset

#     # Create the actual dataset objects.

#     # Solution:
#     # 1. Create the two datasets objects
#     # 2. Duplicate all the dataset membership objects for the two datasets (with the right dataset_id)

#     # 3. Annotate the two datasets with the two annotation props
#     annotation_job_1 = AnnotationJob(
#         input_data=AnnotationInputData(
#             dataset_id=original_dataset_id,
#             files_ids=None,
#             upsert=True,
#             inference_settings=inference_settings_1
#         )
#     )

#     annotation_job_2 = AnnotationJob(
#         input_data=AnnotationInputData(
#             dataset_id=original_dataset_id,
#             files_ids=None,
#             upsert=True,
#             inference_settings=inference_settings_2
#         )
#     )
#     batch_annotate_job_with_checkpoints(
#         identity=identity,
#         job_execution_id=job_execution_id,
#         annotation_job=annotation_job_1,
#         settings=settings,
#         dashboard_db=dashboard_db,
#     )

#     batch_annotate_job_with_checkpoints(
#         identity=identity,
#         job_execution_id=job_execution_id,
#         annotation_job=annotation_job_2,
#         settings=settings,
#         dashboard_db=dashboard_db,
#     )

#     def compute_all_single_file_evals(
#         dataset_1: Dataset,
#         dataset_2: Dataset,
#     ) -> list[SingleFileEval]:

#         single_file_evals: list[SingleFileEval] = []
#         for file_id in dataset_1.file_ids:
#             single_file_evals.append(
#                 SingleFileEval(
#                     file_id=file_id,
#                     dict_1=dataset_1,
#                     dict_2=dataset_2.get_file(file_id),
#                 )
#             )

#         for file_id in dataset_2.file_ids:
#             single_file_evals.append(
#                 SingleFileEval(
#                     file_id=file_id,
#                     dict_1=dataset_2.get_file(file_id),
#                     dict_2=dataset_1,
#                 )
#             )

#         for file_id in dataset_1.file_ids:
#             single_file_evals.append(SingleFileEval(
#                 file_id=file_id,
#                 dict_1=dataset_1.get_file(file_id),
#                 dict_2=dataset_2.get_file(file_id),
#                 schema_id=schema_id,
#                 schema_data_id=schema_data_id,
#                 dataset_membership_id_1=dataset_1.get_file(file_id).id,
#                 dataset_membership_id_2=dataset_2.get_file(file_id).id,
#                 hamming_similarity=compute_dict_difference(
#                     dict_1=dataset_1.get_file(file_id),
#                     dict_2=dataset_2.get_file(file_id),
#                     metric="hamming_similarity"
#                 ),
#                 jaccard_similarity=compute_dict_difference(
#                     dict_1=dataset_1.get_file(file_id),
#                     dict_2=dataset_2.get_file(file_id),
#                     metric="jaccard_similarity"
#                 ),
#                 levenshtein_similarity=compute_dict_difference(
#                     dict_1=dataset_1.get_file(file_id),
#                     dict_2=dataset_2.get_file(file_id),
#                     metric="levenshtein_similarity"
#                 )
#                 )


#         )
#     # Then go through all the entries in the datasets and compute the evaluation metrics
#     compute_all_single_file_evals(
#         dataset_1=dataset_1,
#         dataset_2=dataset_2,
#     )
#     # Return the EvalMetrics object

#     compute_eval_metrics


#     raise NotImplementedError("Not implemented")

#     return eval_metrics
