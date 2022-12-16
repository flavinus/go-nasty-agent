# Nasty Agent : Système de détection des agents mal intentionnés

##  Description

Dans certains centres d’appel, des agents mal intentionnés ont décidé d’éviter certains appels, en feignant un problème de communication.

Voici un exemple de scénario machiavélique utilisé :

* Mettre en attente un long moment le client
* puis
* prétendre de ne pas entendre le client en ayant muté son casque (volume sonore à 0)
* prétendre de ne pas entendre le client sans muter son casque et sans parler (volumesonore n’est pas forcément à 0)
* prétendre de ne pas entendre le client et parler (ex: allo ? Allo ?)
* et attendre que le client raccrocher ou si l’agent à prétendu parler peut également raccrocher

Il existe bien d’autres scénarios pour éviter de devoir converser avec le client.

## Strategy

Une conversation normale se caractérisent par un échange, c'est a dire une succession de prises de parole, silences ou mutes, balancée entre les deux interlocuteurs. 

- Les interlocuteurs vont changer régulièrement d'état au fil de la discussion
- On peut aussi considérer un status de la discussion en fonction du status de chaque channel

### Channel state

A un temps donné, on peut résumer le statut du channel en fonction du volume sonore.

0:MUTE 			volume = 0
1:SILENT 		volume < 10, configurable
2:SPEAKING

RQ: peux etre que le bruit de fond peut etre différents en fonction des interlocuteurs et que la valeur arbitraire choisue posera problème, à voir...


Cela permet de simplifier la lecture de données et de détecter facilement des changements d'état.
Les changement d'état peuvent être considéré comme des évenements et on pourra les compter, et/ou prendre des décisions (on peut prendre en compte la durée).

Ces états permettent d'obtenir une nouvelle séquence, il est aussi possible d'étudier l'ensemble des états d'un channel pour tenter d'en tirer des conclusions (de manière statistique ou en étudiant la form générale).

Par exemple:

    111222222111111122222112222111111111122222222222222222222222221111111222222111211111222111

    111222222110000000000000000000000000000000000000000000000000000000000000000000000000000000


### Discussion state

A un temps donné, on peut définir le status de la conversation en combinant le status des deux channels.

_: SILENT or MUTE
c: CLIENT
a: AGENT
#: COLLISION

A leur tour ces données peuvent être étudiée, notament cela permet d'évaluer la répartition de la parole en comptant simplement les status.

    __cc___aaaaaaaaaaaaaaaaa__cccccc___ccccc_cc___aaaa__a_a_aa_aaaaaa_aaaaa__aaaaaa######____####___#cccccccccccc ...

    __cc____ccccc___cccccccccccc_____________________________________ ...


Une bonne discussion:

* Les prises de paroles se font généralement chacun son tour avec assez peu de moment ou les deux parlent en même temps (collision)
* La répartition du temps de paroles est équilibrée
* Les silences entre les prises de paroles sont supposés être courts

Une mauvaise discussion:

* Mises en attentes avec ou sans mute
* Une répartition des prises de paroles pas du tout équilibrée
* Beaucoup de collisions (ton qui monte ou surdité feinte)
* Coupure brutale de la communication (pas de silence à la fin)

Une fraude:

* Silences ou mute vraiment longs notament en début de conversation
* Très peu de prises de paroles de l'agent
* Un interlocuteur met brutalement fin à la conversation


## Algorithm

Je pense que l'évaluation d'une discussion est une tâche complexe et qu'il est pas très judicieux de vouloir tirer des conclusions en mode tout ou rien en se basant uniquement sur le déclenchement de conditions spécifiques.

Pour moi, une bonne approche serait de s'orienter vers l'utilisation d'un score conversation qui est calculé en fonction de divers indices et de statistiques.
Cette approche n'empêche pas que les indices jugés très pertinents fassent monter très vite le score.
De plus cette approche permet aussi de produire une note à propos de la qualité de la discussion et pas seulement de se prononcer sur les fraudes.

Enfin, le score doit être considéré en fonction de la durée de la conversation, on va donc calculer à la fin un ratio.

* De manière intérative, détecter les changement d'états des channels et leur durée: scorer sur certains évenements (fin de mute)
* Compter les états de la discussion et scorer sur certaines répartitions



## Autres pistes

* un fort niveau sonore aurait pu etre considéré comme un indice ( un interlocuteur qui ne se fait pas entendre à tendance à élever la voix )

* Rechercher des pattern sur un chanel a l'aide regexps, 
    `regexp.MatchString("^[012]{1,50}0+$", statesAgent)`

* en début de conversation, la distribution de la parole peut etre spécifique en fonction du type d'appel: en téléprospection l'agent commence par un speech, en SAV le client explique sa demande => recherche de motifs strès spécifiques

* on aurait pu tenter de produire une séquence des changements d'états

* machine learning, analyse conversationnelle, etc ...

