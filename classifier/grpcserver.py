import grpc
from concurrent import futures
import classifier_pb2
import classifier_pb2_grpc

import numpy as np
import pandas as pd
from sklearn.model_selection import train_test_split
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.svm import OneClassSVM
from sklearn.metrics import make_scorer
from sklearn.model_selection import GridSearchCV


# Load the dataset
data = pd.read_csv("job_application_rejections.csv")
data["label"] = data["label"].apply(lambda x: 1 if x == "reject" else 0)

# Split the data
train_data, test_data = train_test_split(data, test_size=0.2, random_state=42)

# Extract the text
X_train, y_train = train_data["text"], train_data["label"]
X_test, y_test = test_data["text"], test_data["label"]

# Create a TfidfVectorizer instance
vectorizer = TfidfVectorizer()
X_train_tfidf = vectorizer.fit_transform(X_train)
X_test_tfidf = vectorizer.transform(X_test)

# Get the "reject" class training data
X_train_reject = X_train_tfidf[y_train == 1]

# Custom scoring function
def contamination_score(estimator, X):
    y_pred = estimator.predict(X)
    misclassified = (y_pred == -1).sum()
    contamination = misclassified / len(y_pred)
    return -contamination

# Define the hyperparameter search space
param_grid = {
    "kernel": ["linear", "rbf"],
    "nu": [0.01, 0.1, 0.25, 0.5],
    "gamma": [0.01, 0.1, 0.25, 0.5, 1],
}

# Create a OneClassSVM instance with default values
svm = OneClassSVM()

# Create the GridSearchCV instance with cross-validation
grid_search = GridSearchCV(
    svm,
    param_grid,
    scoring=make_scorer(contamination_score, greater_is_better=False),
    cv=5,
    verbose=2,
    n_jobs=-1,
)

# Train the model with the "reject" class only and perform the grid search
grid_search.fit(X_train_reject)

# Get the best hyperparameters
best_params = grid_search.best_params_
print("Best hyperparameters:", best_params)

# Train the One-Class SVM with the best hyperparameters
one_class_svm = OneClassSVM(**best_params)
one_class_svm.fit(X_train_reject)

# Make predictions on the test set
y_test_pred = one_class_svm.predict(X_test_tfidf)

# Transform the predictions from {-1, 1} to {0, 1}
y_test_pred = (y_test_pred + 1) // 2

# Calculate the test accuracy
test_accuracy = np.mean(y_test_pred == y_test)
print("Test accuracy:", test_accuracy)

class ClassifierServicer(classifier_pb2_grpc.ClassifierServicer):
    def ClassifyEmail(self, request, context):
        print("request: ", request.email_text)
        email_text = request.email_text
        # Preprocess and convert the text string to a TF-IDF vector
        text_transformed = vectorizer.transform([email_text])
        
        # Get the prediction from the model
        prediction = one_class_svm.predict(text_transformed)
        prediction = 1 if prediction == 1 else 0  # Convert the prediction to 1 (reject) and 0 (no_reject)
        print("prediction: ", prediction)
        return classifier_pb2.ClassifyResponse(is_rejection=bool(prediction))

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    classifier_pb2_grpc.add_ClassifierServicer_to_server(ClassifierServicer(), server)
    server.add_insecure_port("[::]:50051")
    server.start()
    server.wait_for_termination()

if __name__ == "__main__":
    serve()
