document.addEventListener("DOMContentLoaded", function () {
  document.querySelectorAll(".like-button").forEach(function (button) {
    button.addEventListener("click", function (event) {
      event.preventDefault();
      var postId = this.dataset.postId;
      fetch(`/post/${postId}/like`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      })
        .then((response) => {
          if (response.ok) {
            response.json().then((data) => {
              document.querySelector(`#like-count-${postId}`).textContent =
                data.likes;
            });
          }
        })
        .catch((error) => console.error("Error:", error));
    });
  });

  document.querySelectorAll(".reply-button").forEach(function (button) {
    button.addEventListener("click", function (event) {
      event.preventDefault();
      var postId = this.dataset.postId;
      var replyForm = document.getElementById("reply-form-" + postId);
      replyForm.style.display =
        replyForm.style.display === "block" ? "none" : "block";
    });
  });

  document.querySelectorAll(".like-comment-button").forEach(function (button) {
    button.addEventListener("click", function (event) {
      event.preventDefault();
      var postId = this.dataset.postId;
      var commentId = this.dataset.commentId;
      fetch(`/post/${postId}/comment/${commentId}/like`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      })
        .then((response) => {
          if (response.ok) {
            response.json().then((data) => {
              document.querySelector(
                `#like-count-comment-${commentId}`
              ).textContent = data.likes;
            });
          }
        })
        .catch((error) => console.error("Error:", error));
    });
  });
});
