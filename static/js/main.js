// Attendre que le DOM soit complètement chargé
document.addEventListener("DOMContentLoaded", function () {
  // Ajouter des écouteurs d'événements pour tous les boutons de like des publications
  document.querySelectorAll(".like-button").forEach(function (button) {
    button.addEventListener("click", function (event) {
      event.preventDefault();

      // Récupérer l'ID de la publication à partir de l'attribut data-post-id
      var postId = this.dataset.postId;

      // Envoyer une requête POST pour liker la publication
      fetch(`/post/${postId}/like`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      })
        .then((response) => {
          if (response.ok) {
            // Si la requête est réussie, mettre à jour le compteur de likes
            response.json().then((data) => {
              document.querySelector(`#like-count-${postId}`).textContent =
                data.likes;
            });
          }
        })
        .catch((error) => console.error("Error:", error));
    });
  });

  // Ajouter des écouteurs d'événements pour tous les boutons de réponse
  document.querySelectorAll(".reply-button").forEach(function (button) {
    button.addEventListener("click", function (event) {
      event.preventDefault();

      // Récupérer l'ID de la publication à partir de l'attribut data-post-id
      var postId = this.dataset.postId;

      // Récupérer le formulaire de réponse correspondant et alterner son affichage
      var replyForm = document.getElementById("reply-form-" + postId);
      replyForm.style.display =
        replyForm.style.display === "block" ? "none" : "block";
    });
  });

  // Ajouter des écouteurs d'événements pour tous les boutons de like des commentaires
  document.querySelectorAll(".like-comment-button").forEach(function (button) {
    button.addEventListener("click", function (event) {
      event.preventDefault();

      // Récupérer les ID de la publication et du commentaire à partir des attributs data-post-id et data-comment-id
      var postId = this.dataset.postId;
      var commentId = this.dataset.commentId;

      // Envoyer une requête POST pour liker le commentaire
      fetch(`/post/${postId}/comment/${commentId}/like`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      })
        .then((response) => {
          if (response.ok) {
            // Si la requête est réussie, mettre à jour le compteur de likes du commentaire
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

  // Ajouter des écouteurs d'événements pour tous les boutons de modification des publications
  document.querySelectorAll(".edit-button").forEach(function (button) {
    button.addEventListener("click", function (event) {
      event.preventDefault();

      // Récupérer l'ID de la publication à partir de l'attribut data-post-id
      var postId = this.dataset.postId;

      // Récupérer le formulaire de modification correspondant et alterner son affichage
      var editForm = document.getElementById("edit-form-" + postId);
      editForm.style.display =
        editForm.style.display === "block" ? "none" : "block";
    });
  });
});
